import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

class Icon extends HTMLElement {
	template() {
		return html`
			<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined" rel="stylesheet" />
			<style>
				${css}

				:host {
					display: inline;
				}

				span.material-symbols-outlined {
					${this.size}
					cursor: default;
					vertical-align: bottom;
				}
			</style>

			<span class="material-symbols-outlined">${this.attr("kind")}</span>
		`;
	}
	constructor() {
		super();

		this.attachShadow({mode:'open'});
	}
	onConnected() {
		if ($.get(document, "head link#icons")) {
			return
		}
		$.get(document, "head").appendChild($.create("link", {
			href: "https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined",
			rel: "stylesheet",
			id: "icons",
		}));
	}
	static get observedAttributes() {
		return ["kind", "size"];
	}
	onAttributeChanged(name, old, v) {
		this.render();
	}
	async render() {
		render(this.template(), this.shadowRoot);
	}
	// attribute
	get size() {
		let n = this.attr("size");
		return n ? `font-size: ${n};` : ""
	}
}
customElements.define("c-icon", Icon);

