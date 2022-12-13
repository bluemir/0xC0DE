import * as common from "../common.js"
import * as $ from "../../bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (app) => html`
	<style>
		${common.css}
	</style>
	<slot></slot>
`;

class TabPanel extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}
}
customElements.define("c-tab-panel", TabPanel);
