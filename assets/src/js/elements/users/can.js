import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';
import {css} from "common.js";

function tmpl(can) {
	return html`
		<style>
			${css}

			:host {
			}
			::slotted(*) {
			}
		</style>
		${can ? html`<slot></slot>`:html``}
	`;
}

class CustomElement extends $.CustomElement {
	constructor() {
		super();
		this.can = false;
	}

	async render() {
		render(tmpl.call(this, this.can), this.shadowRoot);
	}

	async onConnected(){
		let verb = this.attr("action");
		let resource = this.attr("resource");

		try {
			let res = await $.request("GET", `/api/v1/can/${verb}/${resource}`);
			this.can = true;
			console.log(res);

			this.render();
		} catch(e) {
			console.error(e);
		}
	}
}
customElements.define("user-can", CustomElement);
