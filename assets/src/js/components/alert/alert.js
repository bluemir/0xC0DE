import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

class Alert extends $.CustomElement {
	template() {
		return html`
			<style>
				${css}

				:host {
					display: block;
					position: fixed;
					top: 3rem;
					right: 1rem;
				}
				article {
					background: white;
					padding: 0.5rem;
					margin: 0.5rem;
				}
			</style>
		`;
	}
	#abort = new AbortController();
	get #signal() {
		return this.#abort.signal;
	}

	constructor() {
		super();
	}

	onConnected() {
		// TODO save last 10 events on session storage
		$.events.on("alert.info", async (evt) => {
			console.log(evt);

			let { message, raw } = evt.detail;

			let $article = $.create("article", {shadow: 2});

			render(html`
				<p>${message}</p>
			`, $article);

			$article.appendTo(this.shadowRoot);

			await $.timeout(3000);

			$article.remove();
		}, {signal:this.#signal});

		$.events.on("alert.warn", async (evt) => {
			let { message, raw } = evt.detail;

			let $article = $.create("article", {shadow: 2});

			render(html`
				<p>${message}</p>
			`,$article);

			$article.appendTo(this.shadowRoot);

			await $.timeout(3000);

			$article.remove();
		}, {signal:this.#signal});

		$.events.on("alert.error", async (evt) => {
			let { message, raw } = evt.detail;

			let $article = $.create("article", {shadow: 2});

			render(html`
				<p>${message}</p>
			`,$article);

			$article.appendTo(this.shadowRoot);

			await $.timeout(3000);

			$article.remove();
		}, {signal:this.#signal});
	}
	onDisconnected() {
		this.#signal.abort();
	}
	async render() {
		render(this.template(), this.shadowRoot);
	}
}
customElements.define("c-alert", Alert);
