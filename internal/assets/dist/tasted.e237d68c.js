const e=document.querySelectorAll(".showable");Array.from(e).forEach((e=>{const n=e.innerHTML,r=r=>{r&&r.preventDefault(),l?(e.innerHTML=n,l=!1):(e.innerHTML="Show",l=!0)};let l=!1;e.addEventListener("click",r),r(null)}));