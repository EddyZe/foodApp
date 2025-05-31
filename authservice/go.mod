module github.com/EddyZe/foodApp/authservice

go 1.24.3

require (
	github.com/EddyZe/foodApp/common v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/getsentry/sentry-go v0.33.0 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace github.com/EddyZe/foodApp/common => ../common
