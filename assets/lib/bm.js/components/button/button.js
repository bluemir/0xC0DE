import * as $ from "../../bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		:host {
			display: inline-block;
			padding: 0.3rem 0.8rem;
		}
		:host(:hover) {
			background: #343434;
		}

		::slotted(*) {
			display: block;
			text-decoration: none;
			color: white;
			white-space: nowrap;
		}
	</style>
	<slot></slot>
`;

class Button extends $.CustomElement {
	constructor() {
		super();
	}
	async render() {
		render(tmpl(this), this.shadow);
	}
}
customElements.define("c-button", Button);
