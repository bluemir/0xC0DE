import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
		}

		global-navigation-bar, ::slotted(*) {
			max-width: 1280px;
			margin: auto;
		}

		header {
			background: var(--blue-alt-700);
			padding: 0.3rem 1rem;
		}
		global-navigation-bar {
			--fg-color: white;
		}
		#hero-contents {
			background: var(--blue-gray-100);
			margin-bottom: 2rem;
			padding: 2rem 1rem;
		}
		main {
			padding: 0 1rem;
		}
	</style>
	<header>
		<global-navigation-bar></global-navigation-bar>
	</header>
	<section id="hero-contents">
		<slot name="hero-contents"></slot>
	</section>
	<main>
		<slot></slot>
	</main>
`;

class CustomElement  extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
}
customElements.define("landing-page", CustomElement);
