import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';
import * as common from "common.js";

class CustomElement extends $.CustomElement {
	template(elem) {
		return html`
			<style>
				${common.css}

				:host {
				}
				::slotted(*) {
					display: inherit;
				}
			</style>
			<slot></slot>
		`;
	}
	static get observedAttributes() {
		return [];
	}
	onAttributeChanged(name, oValue, nValue) {
	}
	constructor() {

	}
	async render() {
		render(this.template(this, this.#params), this.shadowRoot);
	}
	async onConnected () {
	}
	get #params() {
		return $.parsePathParam("/:pathParams");
	}
}
customElements.define("example-element", CustomElement);
