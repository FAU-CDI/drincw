const element = document.querySelectorAll(".showable");
Array.from(element).forEach(element => {
    const content = element.innerHTML
    const placeholder = element.getAttribute('data-placeholder') ?? 'Show'
    const onclick = (event) => {
        if(event) event.preventDefault()
        if (hidden) {
            element.innerHTML = content
            hidden = false
        } else {
            element.innerHTML = placeholder
            hidden = true
        }
    }
    let hidden = false
    
    element.addEventListener('click', onclick)
    onclick(null)
})