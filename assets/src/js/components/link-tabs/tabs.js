import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
			display: flex;
			gap: 1rem;
			margin-bottom: 0.5rem;
		}

		::slotted(.selected) {
			border-bottom: 0.2rem solid var(--green-400);
		}
		::slotted(a) {
			color: black;
			text-decoration: none;
			padding: 0.3rem 0.5rem;
			border-bottom: 0.2rem solid var(--gray-200);
		}
	</style>
	<slot></slot>
`;

class CustomElement extends $.CustomElement {
	constructor() {
		super();
	}
	onConnected() {
		$.all(this, "a").
			forEach(elem => console.log(elem.pathname))
		$.all(this, "a").
			filter((elem) => elem.hasAttribute("exact") ? elem.pathname == location.pathname : location.pathname.startsWith(elem.pathname)).
			filter((elem) => elem.hasAttribute("exact") ? elem.search == location.search: true). // TODO check only pathname?
			forEach(elem => elem.classList.add("selected"));
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
}
customElements.define("c-link-tabs", CustomElement);
