import Msg from './msg.svelte'

function match(routes, url) {
  let uri = url.hash

  if (uri == "") {
    uri = "/"
  } else {
    uri = uri.substring(1).split("?")[0];
  }

  let page = routes.map[uri];
  if (page) {
    return page
  }

  return null
}

async function loadPage(context, routes, url) {
  let res = match(routes, url)

  if (!res) {
    return {
      status: 404,
      render: Msg,
      props: { msg: "404 Not Found." }
    }
  }

  let p = await res.page()
  let params = url.searchParams;

  let tmp = url.hash.substring(1).split("?")
  if (tmp.length > 1) {
    params = new URLSearchParams(tmp[1]);
  }

  let input = {
    url,
    query: params,
    context,
  }

  let resp;
  resp = p.load ? await p.load(input) : {}

  if (resp instanceof Error) {
    return {
      status: 404,
      render: Msg,
      props: { msg: resp.message }
    }
  }

  return {
    status: resp.status || 200,
    redirect: resp.redirect,
    headers: resp.headers,
    render: p.default,
    props: resp.props,
  }
}

function goto(context) {
  return function (href) {
    let url = context.get('url');

    url.update((_url) => {
      let x = new URL(href, _url);
      history.pushState({}, "", x.href)
      return x
    })
  }
}


export { goto, match, loadPage }
