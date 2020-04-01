# Bangunin
Bangunin is a Go-based web app which functions as a call-based alarm. It basically means that it will wake you up by using phone calls. This project is actually part of me studying Golang.

Here is a list of what I have learnt throughout this project.
1. Using `net/http` to build a web app.
2. Using `net/http` to send an HTTP request (POST, GET).
3. Using `regexp` to scrape csrf tokens.
4. Using `time` to schedule the call.
5. Using goroutines

Here is my issues while doing this project.
1. Error handling (I don't know how to properly handle errors, I wrote the same code all the time).
2. Still confused with whether to use exported or unexported identifier.
3. Channels?
4. Routing: "/" route is accessible from any undefined routes that matches "/.*".
5. I don't know how to properly format long lines like Python (is it important in Golang?).
6. I don't know how to properly deploy this web app.

### How does it work?
It works by scraping ![CitCall Demo](https://www.citcall.com/demo/) page. It is illegal though to misuse the demo page. This project is for educational purposes only. I don't encourage you to use this project for other purposes.

## Building
```
git clone https://github.com/p4kl0nc4t/bangunin
cd bangunin
go build .
```

## Usage
```
./bangunin -port=<port>
```

## License
This project is licensed with WTFPL.

## Contribution
Feel free to contribute to this project. Any kind of contribution is really appreciated.
