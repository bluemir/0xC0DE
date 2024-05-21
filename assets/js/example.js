import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';
import {css} from "common.js";

function tmpl() {
	return html`
		<style>
			${css}

			:host {
			}
			::slotted(*) {
			}
		</style>
		<slot></slot>
	`;
}

class CustomElement extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl.call(this), this.shadowRoot);
	}
}
customElements.define("example-element", CustomElement);
