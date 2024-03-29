import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		@import "/static/css/root.css";

		:host {
		}
	</style>
	<span>
		copyright. bluemir
	</span>
`;

class AdminFooter extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}
}
customElements.define("admin-footer", AdminFooter);
