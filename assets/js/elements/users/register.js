import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css, closeDialog} from "common.js";

let tmpl = (elem) => html`
	<style>
		${css}

		:host {
		}

		c-input, section {
			margin: 1rem 0;
		}
	</style>
	<form @submit=${ evt => elem.onSubmit(evt) }>
		<div>
			<c-input label="username"          name="username" type="text"     placeholder="your nickname. eg) bluemir" ></c-input>
		</div>
		<div>
			<c-input label="password"          name="password" type="password" placeholder="min-length: 6" ></c-input>
		</div>
		<div>
			<c-input label="password confirm"  name="confirm"  type="password" placeholder="same as password" ></c-input>
		</div>
		<section>
			<input type="checkbox" id="terms"/>
			<label for="terms"> I read and agree to terms &amp; conditions.</label>
		</section>
		<c-button><button>Create Account</button></c-button>
	</form>
	<dialog @click="${closeDialog}">
		<h1>Login Failed</h1>
	</dialog>
`;

class RegisterForm extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
	async onSubmit(evt) {
		evt.preventDefault();

		try {

			let fd = new FormData($.get(this.shadowRoot, "form"));

			let res = await $.request("POST", `/api/v1/users`, {body:fd});

			location.href = "/"
		} catch(e) {
			console.log(e)
			$.get(this.shadowRoot, "dialog").showModal();
		}
	}
}
customElements.define("register-form", RegisterForm);
