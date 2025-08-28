import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

class Card extends HTMLElement {
	template() {
		return html`
			<style>${css}</style>
			<style>
				:host {
					display: grid;

					padding: 0.5rem;

					box-shadow: 2px 2px 4px 2px rgba(0,0,0,0.5);
				}

				header {
					border-bottom: 1px solid rgba(0,0,0,0.5);
					margin-bottom: 0.5rem;
				}
			</style>
			<section>
				<header>
					<slot name="header"></slot>
				</header>
				<div>
					<slot name="body"></slot>
				</div>
			</section>
		`;
	}
	constructor() {
		super();

		this.attachShadow({mode: 'open'});
	}
	async render() {
		render(this.template(), this.shadowRoot);
	}
}
customElements.define("c-card", Card);
