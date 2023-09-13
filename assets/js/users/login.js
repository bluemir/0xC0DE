import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		@import "/static/css/root.css";

		:host {
		}
		::slotted(*) {

		}
	</style>
	<form>
		<c-input label="username" name="username" type="text"    ></c-input>
		<c-input label="password" name="password" type="password"></c-input>
		<button>Login</button>
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
}
customElements.define("login-form", LoginForm);
