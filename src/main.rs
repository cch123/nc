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
    let html = body.as_str();
    let css_selector = ".article__body-text";

    let document = kuchiki::parse_html().one(html);

    for css_match in document.select(css_selector).unwrap() {
        let as_node = css_match.as_node();

        let mut content = String::new();
        as_node.children().for_each(|n| {
            match n.data() {
                kuchiki::NodeData::Text(t) => {
                    content.push_str(t.clone().into_inner().trim_matches(|c| c == '"'));
                },
                kuchiki::NodeData::Element(_) => {
                    content.push_str(n.text_contents().as_str());
                },
                _ => println!("fuck"),
            }
        });

        println!("{}", content);
    }
}
