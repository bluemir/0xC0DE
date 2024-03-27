import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		@import "/static/css/root.css";

		:host {
		}
		::slotted(*) {
		}
	</style>
	<slot></slot>
`;

class CustomElement extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
}
customElements.define("example-element", CustomElement);
