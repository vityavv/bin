# bin

Basically this is a super-WIP "replacement" for Google Drive. I'd appreciate a better name. I'm making this because I keep trying to use Vim shortcuts in google drive and it is not very great.

## Features

1. Customizable text editor (Ace) with support for Vim/Emacs keybindings, different colorschemes, fonts, and many features of its own.
2. Everything is customizable, including the style of the editor, the style of the pages, and the style of what comes out of the editor (what you render).
3. Image editor using Toast UI!
4. PRESENTATIONS, using the built-in `sent` rendering method: modeled after [suckless' sent](https://tools.suckless.org/sent/).

## Goals

### Editor

1. [x] Vim shortcuts
2. [x] Image editor
3. [ ] Spreadsheet editor
4. [ ] Form editor (export to spreadsheets)

### Files

1. [ ] Move to database
2. [ ] Encryption
3. [ ] Google Drive integration
4. [ ] API

## How to use

1. Install bin by running the command `go get github.com/vityavv/bin`. Make sure you have go installed.
2. Use the command `bin` (I know this is probably a bad name for some) to run the program (make sure the place where go installs stuff is in your PATH.
3. Go to `localhost:8080` in your browser to access the site.
	* You can use a program like NGINX or Apache as a reverse proxy to run bin properly on ports 80 and 443.
4. Create yourself an account, log in, etc. (I'm trying to make it pretty straightfoward)
	* A file will be already in your bin called `.style.css`, which you can edit to change the CSS of every page.
	* Another already created file, `.userScript.js`, will allow you to script the editor (with some pre-put options and functions in place)
	* Finally, a third file, `.renderStyle.css`, will let you style the rendered text.
5. Create a new file or new folder by pressing the buttons on the top. Alternatively, upload a file.
6. Click on a file's name to edit it, or click on a folder's name to enter that folder.
	* You can press the `Back` button at the top of the screen when in a folder to go up one folder.
	* You can also rename the folder by editing the text in the folder's title and pressing the `Rename` button.
7. When editing a file, you can press save to `Save` it (you need to press `Save` when renaming it too), or `Render` to render it with one of the renderers available
	* To write your own, edit `render.go` (that's all you need to edit). It's pretty straight forward, I think.
	* To save an image, you need to press the `Save` button, not the `Load` and `Download` buttons on the right.
