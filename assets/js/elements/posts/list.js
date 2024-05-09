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
	</style>
	${elem.posts.map(post => html`
		<article>${post.message} - ${post.id}</article>
	`)}
`;

class PostList extends $.CustomElement {
	constructor() {
		super();

		this.posts = [];
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
	onConnected() {
		let events = new EventSource("/api/v1/posts/stream");
		events.on("post", evt => {
			console.log(evt);
			let post = JSON.parse(evt.data);
			this.posts = [...this.posts, post]
			this.render();
		})
	}
}
customElements.define("post-list", PostList);
