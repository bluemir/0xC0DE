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
		render(tmpl(this), this.shadow);
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
