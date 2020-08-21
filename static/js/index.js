import $ from "..//lib/web-components/minilib.module.js";
import {html, render, directive} from '../lib/lit-html/lit-html.js';

// TODO will be patched
// See https://github.com/Polymer/lit-html/issues/877
const live = directive((value) => (part) => {
    part.setValue(value);
	part.commit();
});

var tmpl = (app) => html`
<style>
	:host {
	}
</style>
<p>hello world</p>
`;

class CodeIndex extends $.CustomElement {
	constructor() {
		super();

		this.on("connected", () => this.render())
	}

	async render() {
		render(tmpl(this, this.memory), this.shadow);
	}
}
customElements.define("code-index", CodeIndex);
