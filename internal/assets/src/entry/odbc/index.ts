const form = document.getElementById("form") as HTMLFormElement;
const pathbuilder = document.getElementById("pathbuilder") as HTMLTextAreaElement;
const result = document.getElementById("result") as HTMLTextAreaElement;
const resultstatus = document.getElementById("resultstatus") as HTMLSpanElement;
const openResult = document.getElementById("openresult") as HTMLAnchorElement;

const uploadfile = document.getElementById("uploadfile") as HTMLInputElement;
const filestatus = document.getElementById("filestatus") as HTMLSpanElement;
const fakebutton = document.getElementById("fakebutton") as HTMLInputElement;

const selectors = document.getElementById("selectors") as HTMLTextAreaElement;
const uploadSelectors = document.getElementById("uploadselectors") as HTMLInputElement;
const selectorbutton = document.getElementById("selectorbutton") as HTMLButtonElement;
const selectorstatus = document.getElementById("selectorstatus") as HTMLSpanElement;

const selector_load = document.getElementById("selector-load") as HTMLButtonElement;
const selector_empty = document.getElementById("selector-empty") as HTMLButtonElement;

//
// FILE Upload Button
//

function addUploadButton(button: HTMLButtonElement | HTMLInputElement, fileInput: HTMLInputElement, textarea: HTMLTextAreaElement, setStatus: (text: string, success: boolean) => void) {
    const MB = 1048576; // 1 MB in bytes

    button.addEventListener("click", (evt) => {
        evt.preventDefault();
        fileInput.click();
    });

    fileInput.addEventListener("change", async (evt) => {
        const files = fileInput.files;
        if (files === null || files.length !== 1) {
            setStatus("No file(s) selected, nothing loaded. ", false);
            return
        }
    
        const file = files[0];
        if (file.size > MB) {
            setStatus("Unable to load pathbuilder, filesize exceeds 1 MB. ", false);
            return
        }
    
        const text = await file.text();
        textarea.value = text;
        setStatus(`Successfully loaded ${JSON.stringify(file.name)}.`, true);
    })
}

addUploadButton(fakebutton, uploadfile, pathbuilder, setUploadStatus);

function setUploadStatus(text: string, ok: boolean) {
    filestatus.innerHTML = "";
    filestatus.append(document.createTextNode(text));
    filestatus.setAttribute("class", ok ? "ok" : "fail");
}
setUploadStatus("", false);

//
// SELECTORS
//

addUploadButton(selectorbutton, uploadSelectors, selectors, setSelectorStatus);

selector_empty.addEventListener('click', (evt) => {
    evt.preventDefault();

    selectors.value = "";
    setSelectorStatus("Selectors removed. ", true);
})

let loadSelectorCount = 0;


selector_load.addEventListener('click', (evt) => {
    evt.preventDefault();

    loadSelectorCount++;
    handleLoadSelectors(loadSelectorCount);
})

async function handleLoadSelectors(id: number) {
    const [text, ok] = await makeSelectors(pathbuilder.value);
    if (id !== loadSelectorCount) return; // if someone else finished first!

    setSelectorStatus(ok ? "Generated selectors. " : text, ok);
    selectors.value = text;
}

async function makeSelectors(text: string): Promise<[string, boolean]> {
    const response = await fetch("api/v2/makeselectors", {
        method: 'POST',
        body: text,
    })
    return [await response.text(), response.ok];
}


function setSelectorStatus(text: string, ok: boolean) {
    selectorstatus.innerHTML = "";
    selectorstatus.append(document.createTextNode(text));
    selectorstatus.setAttribute("class", ok ? "ok" : "fail");
}
setSelectorStatus("", false);

//
// GENERATING RESULT
//

let counter = 0;
let sendID = 0;

form.addEventListener("submit", (evt) => {
    evt.preventDefault();

    // pick a new id to use for sending!
    sendID++;
    handleSubmit(sendID);
});

async function handleSubmit(id: number) {
    const [text, ok] = await makeODBC(pathbuilder.value, selectors.value.trim());
    if (id !== sendID) return; // if someone else finished first!

    setResultStatus(ok ? "Generated odbc file. " : text, ok);
    setResult(text, ok);
}

async function makeODBC(text: string, selectors: string): Promise<[string, boolean]> {
    const response = await fetch("api/v2/makeodbc", {
        method: 'POST',
        body: JSON.stringify([text, selectors]),
    })
    return [await response.text(), response.ok];
}

function setResultStatus(text: string, ok: boolean) {
    resultstatus.innerHTML = "";
    resultstatus.append(document.createTextNode(text));
    resultstatus.setAttribute("class", ok ? "ok" : "fail");
}
setResultStatus("", false);

function setResult(text: string, ok: boolean) {
    if(!ok) {
        result.value = "";
        openResult.removeAttribute("href");
        return;
    }

    result.value = text;
    
    const href = "data:text/plain;base64," + btoa(text);
    openResult.setAttribute("href", href);
}
setResult("", false); // reset on page load

//
// Processing result
//

result.addEventListener("click", (evt) => {
    evt.preventDefault();
    result.select();
});