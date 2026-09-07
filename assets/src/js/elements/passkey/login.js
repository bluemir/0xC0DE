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
	<button ?disabled=${elem.busy} @click=${() => elem.login()}>
		<slot>Login with Passkey</slot>
	</button>
	${elem.message ? html`<p role="alert">${elem.message}</p>` : ""}
`;

// <passkey-login redirect="/posts"> 는 username 없이 passkey 로 로그인한다.
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

	async login() {
		this.busy = true;
		this.message = "";
		this.render();

		try {
			await passkey.login();

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
customElements.define("passkey-login", CustomElement);
