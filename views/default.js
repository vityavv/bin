editor.setTheme("ace/theme/gruvbox")
// If you don't want to use vim mode, uncomment this line
// editor.setKeyboardHandler("")
// You can see the options here: https://github.com/ajaxorg/ace/wiki/Configuring-Ace
// Or by pressing Ctrl+, while in the editor
editor.setOptions({
	highlightActiveLine: true,
	useSoftTabs: false,
	wrap: "free",
	tabSize: 2,
	showPrintMargin: false,
})
const title = document.getElementById("title")
const modelist = ace.require("ace/ext/modelist")
function detectLang() {
	editor.session.setMode(modelist.getModeForPath(title.value).mode)
}
title.addEventListener("change", detectLang)
detectLang()
