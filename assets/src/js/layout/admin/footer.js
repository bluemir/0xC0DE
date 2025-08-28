import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
		}
	</style>
	<span>
		copyright. bluemir
	</span>
`;

class AdminFooter extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({mode: "open"});
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
}
customElements.define("admin-footer", AdminFooter);
