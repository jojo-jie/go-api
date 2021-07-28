module blog

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/eddycjy/opentracing-gorm v0.0.0-20200209122056-516a807d2182
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.5.0
	github.com/jinzhu/gorm v1.9.16
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/smacker/opentracing-gorm v0.0.0-20181207094635-cd4974441042 // indirect
	github.com/spf13/viper v1.7.1
	github.com/uber/jaeger-client-go v2.29.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/dgrijalva/jwt-go v3.2.0 => github.com/golang-jwt/jwt v3.2.1
