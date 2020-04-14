document.getElementById("newFolder").addEventListener("click", () => {
	window.location.href = window.location.origin + "/newFolder/" + window.location.pathname.slice("/file/".length - 1) + "/" + encodeURIComponent(prompt("Enter a name for the new folder:"))
})
Array.from(document.getElementsByClassName("remove")).forEach(a => {
	a.addEventListener("click", e => {
		if (confirm(`Are you sure you want to delete ${a.id}?`)) window.location.href = window.location.origin + "/remove/{{.File.Path}}/" + a.dataset.filename
	})
})
let FOLDERLIST = "" // for caching
function openMoveDialog(folderList, elementToAppendTo) {
	const folderDiv = document.createElement("div")
	folderDiv.appendChild(document.createTextNode("Choose a folder to move this to:\n"))
	folderDiv.appendChild(document.createElement("br"))
	folderList.split("\n").forEach(folder => {
		if (document.body.dataset.path === folder) {
			return
		}
		const folderLink = document.createElement("a")
		folderLink.href = "javascript:;"
		folderLink.addEventListener("click", move)
		folderLink.innerText = "/" + folder
		folderDiv.appendChild(folderLink)
		folderDiv.appendChild(document.createElement("br"))
	})
	const cancelButton = document.createElement("button")
	cancelButton.classList.add("button")
	cancelButton.addEventListener("click", function() {
		this.parentNode.remove()
	})
	cancelButton.innerText = "Cancel"
	folderDiv.appendChild(cancelButton)
	elementToAppendTo.parentNode.appendChild(folderDiv)
}
function move() {
	const formData = new FormData()
	formData.set("folder", this.innerText)
	fetch(this.parentNode.parentNode.children[2].href.replace("/file/", "/move/"), {
		method: "POST",
		body: formData
	}).then(res => {
		if (!res.ok) {
			res.text().then(err => {
				alert(`Moving failed! - ${res.status}: ${err}`)
			})
			return
		}
		// If it worked, there's a guaranteed redirect
		window.location.href = res.url
	})
}
Array.from(document.getElementsByClassName("move")).forEach(a => {
	a.addEventListener("click", function() {
		if (FOLDERLIST !== "") {
			openMoveDialog(FOLDERLIST, this)
			return
		}
		fetch("/folderList").then(res => {
			if (!res.ok) {
				throw Error("Not ok!")
			}
			return res.text()
		}).then(folderList => {
			FOLDERLIST = folderList
			openMoveDialog(folderList, this)
		})
	})
})
document.querySelector("input#upload").addEventListener("change", e => {
	document.querySelector("#file_chosen").innerText = "File: " + e.srcElement.files[0].name
})
