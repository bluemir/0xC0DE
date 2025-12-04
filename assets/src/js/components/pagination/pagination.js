import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

class CustomElement extends HTMLElement {
	template() {
		return html`
			<style>
				${css}
			</style>
			<button enhanced @click="${evt => this.prev(evt)}">Prev</button>
			<span>${this.attr("page")}</span>
			<button enhanced @click="${evt => this.next(evt)}">Next</button>
		`;
	}

	constructor() {
		super();

		this.attachShadow({mode: "open"});
	}
	render() {
		render(this.template(), this.shadowRoot);
	}
	static get observedAttributes() {
		return ["page"];
	}
	onAttributeChanged(name, oValue, nValue) {
		this.render();
	}

	//
	set page(v) {
		// FIXME validate
		v = parseInt(v) || 1;
		v = v < 1? 1: v;

		this.attr("page", v)

		this.fireEvent("change", v)

		this.render();
	}
	get page() {
		return parseInt(this.attr("page")) || 1;
	}

	//
	prev(evt) {

		this.page -= 1;
		if (this.page < 1){
			this.page = 1
		}
	}
	next(evt) {
		this.page += 1
	}

}
customElements.define("c-pagination", CustomElement);
