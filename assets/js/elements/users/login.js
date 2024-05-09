import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
		}
		::slotted(*) {
		}
		c-input {
			margin: 1rem 0;
		}
	</style>
	<form @submit="${evt => elem.onSubmit(evt)}">
		<div>
			<c-input label="username" name="username" type="text"    ></c-input>
		</div>
		<div>
			<c-input label="password" name="password" type="password"></c-input>
		</div>
		<div>
			<c-button><button>Login</button></c-button>
		</div>
	</form>
	<dialog>
		<h1>Register Failed</h1>
	</dialog>
`;

class CustomElement extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
	async onSubmit(evt) {
		evt.preventDefault();

		let fd = new FormData($.get(this.shadowRoot, "form"));

		let res = await $.request("POST", `/api/v1/login`, {body:fd});

		location.href = "/posts"
	}
}
customElements.define("login-form", CustomElement);
