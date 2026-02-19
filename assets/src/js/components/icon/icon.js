import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "@/common.js";

class Icon extends HTMLElement {
	template() {
		if (this.fa) {
			let prefix = this.fa === "brand" ? "fa-brands" : "fa-solid";
			return html`
				<style>
					${css}

					:host {
						display: inline-flex;
					}

					i {
						${this.size}
					}
				</style>

				<i class="${prefix} fa-${this.attr("kind")}"></i>
			`;
		}

		return html`
			<style>
				${css}

				:host {
					display: inline-flex;
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
	static get observedAttributes() {
		return ["kind", "size", "fa"];
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
	get fa() {
		return this.attr("fa");
	}
}
customElements.define("c-icon", Icon);
