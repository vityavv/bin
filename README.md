# bin

Basically this is a super-WIP "replacement" for Google Drive. I'd appreciate a better name. I'm making this because I keep trying to use Vim shortcuts in google drive and it is not very great.

## Goals

1. [x] Vim shortcuts in the editor
	1. [x] Make the escape key work!
2. [x] Customizable editor (probably goona use Ace for this)
3. [ ] Render Markdown or even LaTeX
4. [ ] Move files to a database
5. [ ] File encryption
6. [ ] Integrate with Google Drive

## How to use

1. `go get github.com/vityavv/bin`
2. `bin` (I know this is probably a bad name for some)
3. go to `localhost:8080` in your browser
4. Create yourself an account, log in, etc. (I'm trying to make it pretty straightfoward)
5. A file will be already in your bin called `.style.css`, which you can edit to change the CSS of every page.
6. Another already created file, `.userScript.js`, will allow you to script the editor (with some pre-put options and functions in place)
