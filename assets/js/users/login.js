import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		@import "/static/css/root.css";

		:host {
		}
		::slotted(*) {

		}
		c-input {
			margin: 1rem 0;
		}
	</style>
	<form @submit="${evt => elem.onSubmit(evt)}">
		<c-input label="username" name="username" type="text"    ></c-input>
		<c-input label="password" name="password" type="password"></c-input>
		<c-button><button>Login</button></c-button>
	</form>
`;

class LoginForm extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}
	async onSubmit(evt) {
		evt.preventDefault();

		let fd = new FormData($.get(this.shadowRoot, "form"));

		let res = await $.request("POST", `/api/v1/login`, {body:fd});

		location.href = "/posts"
	}
}
customElements.define("login-form", LoginForm);