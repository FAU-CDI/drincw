const element = document.querySelectorAll(".showable");
Array.from(element).forEach(element => {
    const content = element.innerHTML
    const onclick = (event) => {
        if(event) event.preventDefault()
        if (hidden) {
            element.innerHTML = content
            hidden = false
        } else {
            element.innerHTML = "Show"
            hidden = true
        }
    }
    let hidden = false
    
    element.addEventListener('click', onclick)
    onclick(null)
})