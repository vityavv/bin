window.onload = main

function main() {
	const slides = Array.from(document.getElementsByClassName("slide"))
	slides.forEach((e,i) => {
		textFit(e, {maxFontSize: 200})
		if (i !== 0) e.style.display = "none"
		e.addEventListener("keydown", keydown)
	})
	slides[0].focus()
}

function keydown(ev) {
	const index = Array.prototype.indexOf.call(this.parentNode.childNodes, this)
	if (ev.code === "ArrowRight") {
		if (index !== this.parentNode.childNodes.length - 1) {
			const newSlide = this.parentNode.childNodes[index + 1]
			newSlide.style.display = "flex"
			this.style.display = "none"
			newSlide.focus()
		}
	} else if (ev.code === "ArrowLeft") {
		if (index !== 0) {
			const newSlide = this.parentNode.childNodes[index - 1]
			newSlide.style.display = "flex"
			this.style.display = "none"
			newSlide.focus()
		}
	}
}
