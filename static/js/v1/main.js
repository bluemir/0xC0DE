import * as $ from "/static/lib/web-components/minilib.module.js";
import {html, render} from '/lib/lit-html/lit-html.js';
//import {html, render} from 'lit-html';

var tmpl = (app) => html`
<style>
	:host {
	}
</style>
<p>hello world</p>
`;

class CodeMain extends $.CustomElement {
	constructor() {
		super();

		this.on("connected", () => this.render())
	}

	async render() {
		render(tmpl(this), this.shadow);
	}
}
customElements.define("code-main", CodeMain);
