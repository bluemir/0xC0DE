import * as $ from "../../bm.module.js";
import {html, render} from 'lit-html';
import {css} from "../common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
			display: inline-block;
			padding: 0.3rem 0.8rem;

			color: var(--button-fg-color, black);
			background: var(--button-bg-color, white);
			border: 1px solid var(--button-border-color, #343434);
		}
		:host(:hover) {
			color: var(--button-hover-fg-color, white);
			background: var(--button-hover-bg-color, #343434);
			border: 1px solid var(--button--hover-border-color, #343434);
		}
		::slotted(*) {
			color: inherit;
			--fg-color: reset;
			text-decoration: none;
			white-space: nowrap;
		}
		::slotted(a:hover) {
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
