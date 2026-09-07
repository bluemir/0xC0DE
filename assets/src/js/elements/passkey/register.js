import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css, errorMessage} from "@/common.js";
import * as passkey from "@/passkey.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
			display: contents;
		}
		p[role="alert"] {
			color: var(--red-600);
		}
	</style>
	<button type="button" ?disabled=${elem.busy} @click=${() => elem.register()}>
		<slot>Register with Passkey</slot>
	</button>
	${elem.message ? html`<p role="alert">${elem.message}</p>` : ""}
`;

// <passkey-register username-from="[name=username]" redirect="/"> 는
// passkey 만으로 계정을 만든다. username-from 이 없으면 로그인한 계정에 passkey 를 추가한다.
class CustomElement extends HTMLElement {
	constructor() {
		super();

		this.attachShadow({mode: "open"});

		this.busy = false;
		this.message = "";
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}

	async onConnected() {
		if (!passkey.supported()) {
			this.hidden = true;
		}
	}

	async register() {
		this.busy = true;
		this.message = "";
		this.render();

		try {
			let selector = this.attr("username-from");
			let username = selector ? $.get(selector).value : "";

			await passkey.register({username, label: this.attr("label") || ""});

			location.href = this.attr("redirect") || "/";
		} catch(e) {
			console.error(e);

			this.message = errorMessage(e);
		} finally {
			this.busy = false;
			this.render();
		}
	}
}
customElements.define("passkey-register", CustomElement);
