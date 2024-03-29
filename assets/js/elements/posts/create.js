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
	<form @submit="${evt => elem.onSubmit(evt)}">
		<c-input label="message" name="message"></c-input>
		<c-button><button>Send</button></c-button>
	</form>
`;

class PostCreate extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}

	async onSubmit(evt) {
		evt.preventDefault();

		let $form = evt.target;
		let fd = new FormData($form);

		let res = await $.request("POST", `/api/v1/posts`, {body:fd});

		console.log(res)

		$form.reset();
	}
}
customElements.define("post-create", PostCreate);
