
build:
		mkdir build
		go build -o build/echo

zip: build
		rm -f echo.zip
		cd build; zip echo.zip echo
