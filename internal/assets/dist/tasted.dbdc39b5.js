!function(){const e=document.querySelectorAll(".showable");Array.from(e).forEach((e=>{const n=e.innerHTML;var t;const r=null!==(t=e.getAttribute("data-placeholder"))&&void 0!==t?t:"Show",l=t=>{t&&t.preventDefault(),o?(e.innerHTML=n,o=!1):(e.innerHTML=r,o=!0)};let o=!1;e.addEventListener("click",l),l(null)}))}();