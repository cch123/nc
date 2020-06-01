// https://www.economist.com/finance-and-economics/2020/05/30/the-pandemic-could-lead-statisticians-to-change-how-they-estimate-gdp
// https://www.economist.com/printedition/2020-05-30
use kuchiki::traits::*;
use std::time::Duration;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let c = reqwest::Client::builder()
        .connection_verbose(true)
        .timeout(Duration::from_secs(30))
        .build()?;

    let body = c
        .get("https://www.economist.com/finance-and-economics/2020/05/30/the-pandemic-could-lead-statisticians-to-change-how-they-estimate-gdp")
        //.get("https://www.economist.com/printedition/2020-05-30")
        .send()
        .await?
        .text()
        .await?;
    //dbg!(body);
    parse(body.clone());
    Ok(())
}

fn parse(body: String) {
    let html = r"
        <DOCTYPE html>
        <html>
        <head></head>
        <body>
            <h1>Example</h1>
            <p class='foo'>Hello, <li>fuck</li>world!</p>
            <p class='foo'>I love HTML</p>
        </body>
        </html>
    ";
    let html = body.as_str();
    //class="article__body-text"
    let css_selector = ".article__body-text";

    let document = kuchiki::parse_html().one(html);

    for css_match in document.select(css_selector).unwrap() {
        // css_match is a NodeDataRef, but most of the interesting methods are
        // on NodeRef. Let's get the underlying NodeRef.
        let as_node = css_match.as_node();

        // In this example, as_node represents an HTML node like
        //
        //   <p class='foo'>Hello world!</p>"
        //
        // Which is distinct from just 'Hello world!'. To get rid of that <p>
        // tag, we're going to get each element's first child, which will be
        // a "text" node.
        //
        // There are other kinds of nodes, of course. The possibilities are all
        // listed in the `NodeData` enum in this crate.
        as_node.children().for_each(|n| {
            match n.data() {
                kuchiki::NodeData::Text(t) => println!("{:?}", t.clone().into_inner()),
                kuchiki::NodeData::Element(e) => println!("{:?}", n.first_child().unwrap().as_text().unwrap().clone().into_inner()),
                _ => println!("fuck"),
            }
        });

        // Let's get the actual text in this text node. A text node wraps around
        // a RefCell<String>, so we need to call borrow() to get a &str out.
        //let text = text_node.as_text().unwrap();//.borrow();

        // Prints:
        //
        //  "Hello, world!"
        //  "I love HTML"
    }
}
