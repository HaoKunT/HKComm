/* package hkcomm
需要实现的功能
1. 两个终端的即时通讯
2. 群功能
3. 好友功能
4. 文件传递功能
5. 历史记录功能
*/
// 实现逻辑
/*
接收和发送在不同的gorouties里面，中间采用通道进行通讯
问题：
在udp服务器中如何分辨是哪个用户发来的消息
解决方案：
1. udp发送时携带用户信息：携带用户id号，int值，只需要4字节即可搞定
问题：
反向发送数据如何保证数据接收到了？
解决方案：
1. 使用TCP，放弃UDP方案
2. 对UDP做一定的控制
3. 采用同步机制，即隔一段时间客户端向服务器发送同步指令
4. 通讯整体采用websocket方案

最终思路：
采用websocket方案进行开发，认证使用session进行认证即可

通道中需要包含发送的对象和发送的消息两部分数据

一个map数据结构用于维持已有的所有连接上的连接，其中map的key为用户名称，表明此连接的客户端对应的用户是哪个，
value是对应的连接列表，因为可能一个用户有多个客户端在线

当websocket建立之后注册接收到消息的回调，回调函数中读取接收到的消息，其中包含消息发送的对象（用户名）
解析消息结构完成后将消息传递给通道，由发送消息的函数根据用户名找到用户对应的连接，并将消息在此websocket上发送出去。
若接收消息的用户没有在线，则将消息暂存在内存中，等用户上线后再将消息发送出去

群聊功能：
将所有处于一个群内的用户加入至同一room内，发送至群聊的消息将向整个群聊发送
需要设定数据库表
问题待解决：
会向自己发送吗？（貌似会）
解决方案：
让前端丢弃由自己向自己发送的消息

好友功能：
设定好友关系列表（数据库），客户端本地缓存好友信息，如果是第一次登录则向服务器请求好友信息（http）
问题：
添加好友时的添加消息应该进入队列，确认添加消息前端接收到之后再从队列中删除
同理聊天功能也需要在对方下线时保证对方上线后能收到消息，群消息也需要在上线后能接收到离线时的群消息

文件传输功能
离线传输和在线传输
小文件默认直接离线传输，大文件默认在线传输，其中可切换
离线传输比较简单，向服务器发送文件，文件标明双方的用户id，将文件存储至数据库（http），并向接收方发送websocket消息，接收方发送接收文件请求，由服务器发送文件
在线传输则首先发送一个文件传输请求，请求中需要标明双方用户，向文件接收方发送websocket消息
问题：
采用类似内网穿透的方式直接让双方接收文件？
可能的解决方案：
写一个，或者用一个p2p通讯框架

消息队列：
生产端是将需要发送至用户的消息数据先推送至消息队列，直到用户上线再将数据发送至用户
消息队列使用RabbitMQ，使用官方推荐的GO第三方包进行开发
测试，在内网环境下，单线程不间断推送简短消息，大约1.5w每秒

RabbitMQ:
为每个离线用户创建离线队列，统合在一个exchange下面就可以了，利用topics的方式
用户上线后在上线中回调消费队列中的相应用户离线消息，发送至用户
当用户长时间不上线时，将用户的离线消息存储至硬盘，用户上线时，将硬盘和队列中的消息整合发送至用户
*/
package hkcomm

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
	"github.com/streadway/amqp"
	"log"
	"sync/atomic"
)

func setUpWebsocket(app *iris.Application) {
	ws := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	ws.OnConnection(func(craw websocket.Connection) {
		// 首先初始化此连接信息，根据连接的session判断是什么用户
		// 放弃框架提供的群的概念（这个概念只能向在线用户发送数据）
		c := &connection{
			craw,
			0,
		}
		s := sess.Start(craw.Context())
		if auth, _ := s.GetBoolean("authenticated"); !auth {
			checkError(c.Emit("auth", "403 Forbidden"))
			checkError(c.Disconnect())
			return
		}
		userId, err := s.GetInt("userid")
		if userId < 0 {
			checkError(fmt.Errorf("Invalid userid: %d", userId))
			return
		}
		if checkError(err) != nil {
			return
		}

		// 将连接加入连接池
		var user *User
		user.ID = uint(userId)
		c.senderID = user.ID
		// 连接加入连接池之前，先将最新的100条信息发送过去

		conns = GetConns()
		conns.Lock()
		conns.conn[user.ID] = append(conns.conn[user.ID], c)
		conns.Unlock()

		//关闭连接时检查连接池
		c.OnDisconnect(func() {
			conns = GetConns()
			conns.Lock()
			// 如果此连接是最后一个连接，则删除整个键，不是则删除对应连接
			if len(conns.conn[c.senderID]) == 1 {
				delete(conns.conn, c.senderID)
			} else {
				index := -1
				for i, conn := range conns.conn[c.senderID] {
					if conn == c {
						index = i
						break
					}
				}
				if index == -1 {
					checkError(fmt.Errorf("wrong connection %v", c))
					return
				}
				conns.conn[c.senderID] = append(conns.conn[c.senderID][:index], conns.conn[c.senderID][index+1:]...)
			}
			conns.Unlock()
		})

		// 这里需要绑定接收到消息的函数
		c.On("group", func(cd communicationData) {
			cd.connection = c
			// 进入处理通道
			cd.Isgroup = true
			cd.Isfile = false
			msgInCh <- cd
		})
		c.On("user", func(cd communicationData) {
			cd.connection = c
			// 进入处理通道
			cd.Isgroup = false
			cd.Isfile = false
			msgInCh <- cd
		})
		c.On("file", func(cd communicationData) {
			cd.connection = c
			// 进入处理通道
			cd.Isfile = true
			msgInCh <- cd
		})
	})

}

func msgCh() {
	for i := 0; i < int(msgNumber); i++ {
		go func() {
			for msgIn := range msgInCh {
				// 根据connection获取发送者ID，与包不符合则丢弃
				if msgIn.connection.senderID != msgIn.From {
					checkError(fmt.Errorf("connection sender (id:%d) mismatch the sender (id:%d)", msgIn.connection.senderID, msgIn.From))
					continue
				}
				// 给消息附上唯一的ID
				msgIn.generateID()
				// 判断是否有这个成员，避免出现精心构造好的包
				_, err := msgIn.GetSender(false)
				if checkError(err) != nil {
					continue
				}
				// 如果是群消息
				if msgIn.Isgroup {
					// 首先判断发送用户是否在这个群内，避免出现精心构造好的包
					user, err := msgIn.GetSender(true)
					if checkError(err) != nil {
						continue
					}
					_, err = contain(msgIn.To, user.Groups)
					if err != nil {
						checkError(fmt.Errorf("user %s not in group %d: %s", user.UserName, msgIn.To, err))
						continue
					}
				} else {
					// 如果是发给个人的消息
					// 判断是否有接收者
					_, err = msgIn.GetReceiver(false)
					if checkError(err) != nil {
						continue
					}
				}
				// 这里对文件消息，如果是文件消息则将之放入缓存（暂时不向接收端发送，等待客户端文件发送完毕后再向接收端发送）
				if msgIn.Isfile {
					filecache = GetFilecache()
					filecache.Lock()
					filecache.cache[msgIn.ID] = &msgIn
					filecache.Unlock()
					if err := msgIn.connection.Emit("file-receipt", receipt{
						msgIn.ID,
						msgIn.Sync,
						200,
						"next",
					}); err != nil {
						checkError(err)
					}
				} else {
					// 不是文件消息
					// 将消息放至消息队列中
					msgOutCh <- msgIn
				}
			}
		}()
		go func() {
			for msgOut := range msgOutCh {
				// 前期判断检查全部完成
					// 将消息送入消息队列，首先序列化对象为json
					msgJson, err := json.Marshal(msgOut)
					if checkError(err) != nil {
						continue
					}
					// publish至HKComm这个exchange
					err = chAMQP.Publish(
						"hkcomm",
						"hkcomm.msg",
						false,
						false,
						amqp.Publishing{
							ContentType: "text/plain",
							Body: msgJson,
						})
					if checkError(err) != nil {
						continue
					}
					// 将confirm同步位加1
					atomic.AddUint64(&syncTag, 1)
					// 将离线同步结构置入离线确认队列中
					Confirm <- &unConfirmTag{
						msgOut.Sync,
						msgOut.connection,
						syncTag,
						&msgOut,
					}

			}
		}()
		// 这个线程用于消费队列中的消息
		go func() {
			q, err := chAMQP.QueueDeclare(
				"",
				false,
				false,
				true,
				false,
				nil)
			if checkError(err) != nil {
				log.Fatalf("QueueDeclare error: %s", err)
			}
			err = chAMQP.QueueBind(
				q.Name,
				"hkcomm.msg",
				"hkcomm",
				false,
				nil)
			if checkError(err) != nil {
				log.Fatalf("QueueBind error: %s", err)
			}
			msgs, err := chAMQP.Consume(
				q.Name,
				"",
				false,
				false,
				false,
				false,
				nil)
			if checkError(err) != nil {
				log.Fatalf("Consume error: %s", err)
			}
			for msg := range msgs {
				var msgOut communicationData
				err = json.Unmarshal(msg.Body, &msgOut)
				if checkError(err) != nil {
					msg.Ack(false)
					continue
				}
				if msgOut.Isgroup {
					// 如果是发送给群的消息
					// 获取群内所有成员
					var group Group
					group.ID = msgOut.To
					group.GetUInfo()
					// 向群内所有用户发送消息
					conns := GetConns()
					conns.RLock()
					for _, user := range group.Users {
						sendMsg(user.ID, &msgOut)
					}
				} else {
					sendMsg(msgOut.To, &msgOut)
				}
				msg.Ack(false)
			}
		}()
	}
	// 这个go程用来确认消息
	go func() {
		confirmTemp := make([]*unConfirmTag, 0)
		for {
			select {
			case confirm := <- Confirm:
				confirmTemp = append(confirmTemp, confirm)
			case confirmTag := <- confirms:
				// 似乎有点小bug，在连接关闭的时候会有默认的tag出来，还出来很多个，这里补救一下
				if confirmTag.DeliveryTag == 0 {
					continue
				}
				index := -1
				for i, con := range confirmTemp {
					if con.confirmTag == confirmTag.DeliveryTag {
						index = i
						break
					}
				}
				if index == -1 {
					checkError(fmt.Errorf("no Confirmed: %d", confirmTag.DeliveryTag))
					continue
				}
				// 这里有个bug，在文件消息收到的不是确认的时候，将会导致消息二次进入Out通道，文件与否将二次判断
				// 解决方法，文件等待在In通道做
				if !confirmTag.Ack {
					msgOutCh <- *confirmTemp[index].data
					continue
				}
				if confirmTemp[index].data.SaveMsg() != nil {
					msgOutCh <- *confirmTemp[index].data
					continue
				}
				if err := confirmTemp[index].sender.Emit("receipt", receipt{
					confirmTemp[index].data.ID,
					confirmTemp[index].sync,
					200,
					"ok",
				}); err != nil {
					checkError(fmt.Errorf("send receipt error: %s", err))
					continue
				}
				confirmTemp = append(confirmTemp[:index], confirmTemp[index+1:]...)
			}
		}
	}()
}


func contain(id uint, groups []*Group) (*Group, error) {
	for _, group := range groups {
		if group.ID == id {
			return group, nil
		}
	}
	return nil, checkError(fmt.Errorf("contain error: group %d not in %v", id, groups))
}

