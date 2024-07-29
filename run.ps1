# Set environment variables
$env:VOSK_PATH = "C:\Users\NotMyPc\Documents\Projects\go\discordBot\nlp\vosk-win64-0.3.45"
$env:CGO_CPPFLAGS = "-I $env:VOSK_PATH"
$env:CGO_LDFLAGS = "-L $env:VOSK_PATH -lvosk"
$env:PATH += ";C:\Users\NotMyPc\Documents\Projects\go\discordBot\nlp\vosk-win64-0.3.45"

# Run the Go program
go run main.go
