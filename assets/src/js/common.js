import * as $ from "bm.js/bm.module.js";

let rev = $.get(`head script[type="importmap"]`).attr("rev");

export let css = `
@import url("/static/${rev}/css/element.css");
`;


export function closeDialog(evt) {
	if (evt.target.nodeName === 'DIALOG') {
		evt.target.close();
	}
}
