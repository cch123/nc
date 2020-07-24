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
        //.get("https://www.economist.com/finance-and-economics/2020/05/30/the-pandemic-could-lead-statisticians-to-change-how-they-estimate-gdp")
        .get("https://www.economist.com/printedition/2020-05-30")
        .send()
        .await?
        .text()
        .await?;

    //parse(body.clone());
    get_table(body.clone());
    Ok(())
}

fn get_table(body : String) {
    let html = body.as_str();
    let css_selector = ".list__item";

    let doc = kuchiki::parse_html().one(html);

    for list_item in doc.select(css_selector).unwrap() {
        let list_item_node= list_item.as_node();
        let mut title = String::new();
        list_item_node.select(".list__title").unwrap().for_each(|t|{
            let text = t.text_contents().clone();
            title.push_str(text.as_str());
        });
        dbg!(title);
        let mut item_arr = Vec::new();
        list_item_node.select("a").unwrap().for_each(|t|{
            let text = t.text_contents().clone();
            let link = t.attributes.borrow().get("href").unwrap().to_string();
            item_arr.push(text.clone());
            item_arr.push(link.to_string());
        });
        dbg!(item_arr);
    }
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

        println!("{}\n", content);
    }
}
