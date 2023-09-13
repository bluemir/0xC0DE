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
		<input label="message" name="message" />
		<button>Send</button>
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

		let fd = new FormData($.get(this.shadowRoot, "form"));

		let res = await $.request("POST", `/api/v1/posts`, {body:fd});

		console.log(res)
	}
}
customElements.define("post-create", PostCreate);
