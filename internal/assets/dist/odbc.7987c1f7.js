const e=document.getElementById("form"),t=document.getElementById("pathbuilder"),n=document.getElementById("result"),a=document.getElementById("resultstatus"),d=document.getElementById("openresult"),i=document.getElementById("uploadfile"),l=document.getElementById("filestatus");document.getElementById("fakebutton").addEventListener("click",(e=>{e.preventDefault(),i.click()}));function o(e,t){l.innerHTML="",l.append(document.createTextNode(e)),l.setAttribute("class",t?"ok":"fail")}i.addEventListener("change",(async e=>{const n=i.files;if(null===n||1!==n.length)return void o("No file(s) selected, nothing loaded. ",!1);const a=n[0];if(a.size>1048576)return void o("Unable to load pathbuilder, filesize exceeds 1 MB. ",!1);const d=await a.text();t.value=d,o(`Successfully loaded ${JSON.stringify(a.name)}.`,!0)})),o("",!1);let c=0;function u(e,t){a.innerHTML="",a.append(document.createTextNode(e)),a.setAttribute("class",t?"ok":"fail")}function s(e,t){if(!t)return n.value="",void d.removeAttribute("href");n.value=e;const a="data:text/plain;base64,"+btoa(e);d.setAttribute("href",a)}e.addEventListener("submit",(e=>{e.preventDefault(),c++,async function(e){const[n,a]=await async function(e){const t=await fetch("api/v1/makeodbc",{method:"POST",body:e});return[await t.text(),t.ok]}(t.value);if(e!==c)return;u(a?"Generated odbc file. ":n,a),s(n,a)}(c)})),u("",!1),s("",!1),n.addEventListener("click",(e=>{e.preventDefault(),n.select()}));