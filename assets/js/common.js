export let css = `
@import url("/static/css/element.css");
`;


export function closeDialog(evt) {
	if (evt.target.nodeName === 'DIALOG') {
		evt.target.close();
	}
}
