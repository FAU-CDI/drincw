// global typescript for everything
const html_content = process.env.LEGAL_HTML;
if (html_content) {
  const footer = document.createElement("footer");
  footer.innerHTML = html_content;
  document.body.append(footer);
}
