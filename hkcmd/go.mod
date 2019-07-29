module hkcmd

go 1.12

require (
	github.com/Unknwon/goconfig v0.0.0-20190425194916-3dba17dd7b9e
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jinzhu/gorm v1.9.10
	github.com/kataras/iris v11.1.1+incompatible
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	gopkg.in/go-playground/validator.v9 v9.29.1
)

require (
	github.com/spf13/cobra v0.0.5
	hkcomm v0.0.0
)

replace hkcomm => "D://learning/code/HKComm/hkcmd/HKComm"
