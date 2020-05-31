// https://www.economist.com/finance-and-economics/2020/05/30/the-pandemic-could-lead-statisticians-to-change-how-they-estimate-gdp
// https://www.economist.com/printedition/2020-05-30
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
    dbg!(body);
    Ok(())
}
