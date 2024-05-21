import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';
import {css} from "common.js";

let tmpl = () => html`
	<style>
		${css}

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
		render(tmpl.call(this), this.shadowRoot);
	}

	async onConnected(){
		let verb = this.attr("action");
		let resource = this.attr("resource");

		let res = await $.request("GET", `/api/v1/can/${verb}/${resource}`);

	}
}
customElements.define("user-can", CustomElement);
