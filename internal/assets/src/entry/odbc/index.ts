const form = document.getElementById("form") as HTMLFormElement;
const pathbuilder = document.getElementById("pathbuilder") as HTMLTextAreaElement;
const result = document.getElementById("result") as HTMLTextAreaElement;
const resultstatus = document.getElementById("resultstatus") as HTMLSpanElement;
const openResult = document.getElementById("openresult") as HTMLAnchorElement;

const uploadfile = document.getElementById("uploadfile") as HTMLInputElement;
const filestatus = document.getElementById("filestatus") as HTMLSpanElement;
const fakebutton = document.getElementById("fakebutton") as HTMLInputElement;


//
// FILE Upload Button
//

fakebutton.addEventListener("click", (evt) => {
    evt.preventDefault();
    uploadfile.click();
});

const MB = 1048576;
uploadfile.addEventListener("change", async (evt) => {
    const files = uploadfile.files;
    if (files === null || files.length !== 1) {
        setUploadStatus("No file(s) selected, nothing loaded. ", false);
        return
    }

    const file = files[0];
    if (file.size > MB) {
        setUploadStatus("Unable to load pathbuilder, filesize exceeds 1 MB. ", false);
        return
    }

    const text = await file.text();
    pathbuilder.value = text;
    setUploadStatus(`Successfully loaded ${JSON.stringify(file.name)}.`, true);
})

function setUploadStatus(text: string, ok: boolean) {
    filestatus.innerHTML = "";
    filestatus.append(document.createTextNode(text));
    filestatus.setAttribute("class", ok ? "ok" : "fail");
}
setUploadStatus("", false);

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
    const [text, ok] = await makeODBC(pathbuilder.value);
    if (id !== sendID) return; // if someone else finished first!

    setResultStatus(ok ? "Generated odbc file. " : text, ok);
    setResult(text, ok);
}

async function makeODBC(text: string): Promise<[string, boolean]> {
    const response = await fetch("api/v1/makeodbc", {
        method: 'POST',
        body: text,
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