const form = document.getElementById("form");
const textarea = form.querySelector("textarea");
const result = document.getElementById("result");

let counter = 0;

let sendID = 0;

form.addEventListener("submit", (evt) => {
    evt.preventDefault();

    // pick a new id to use for sending!
    sendID++;
    handleSubmit(sendID);
});

async function handleSubmit(id) {
    const pathbuilder = textarea.value;
    const odbcResult = await makeODBC(pathbuilder);
    setResult(id, odbcResult[0], odbcResult[1]);
}

function setResult(id, ok, text) {
    // some other request finished first
    // so don't display the old one!
    if (id !== sendID) return;

    // clear out old response!
    result.innerHTML = "";


    // we didn't get a valid response!
    if (!ok) {
        result.appendChild(document.createTextNode(text));
        return;
    }

    // we did get a valid response!
    const code = document.createElement("code");
    const pre = document.createElement("pre");
    code.appendChild(pre);
    pre.appendChild(document.createTextNode(text));
    result.appendChild(code);
}

/**
 * 
 * @param {string} odbc ODBC code to send 
 * @returns a pair (ok: bool, error_or_result: string)
 */
async function makeODBC(text) {
    const response = await fetch("api/v1/makeodbc", {
        method: 'POST',
        body: text,
    })
    return [response.ok, await response.text()];
}