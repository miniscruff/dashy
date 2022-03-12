function createDiv(classNames, textContent) {
  const div = document.createElement("div");

  if (Array.isArray(classNames)) {
    div.classList.add(...classNames);
  } else {
    div.className = classNames;
  }

  if (textContent) {
    div.textContent = textContent;
  }

  return div
}

function addCard(headerText, cssStyles) {
  const cardDiv = createDiv("card");
  const headerDiv = createDiv("card-header", headerText);

  const layoutStyles = [...cssStyles, "card-layout"]
  const layoutDiv = createDiv(layoutStyles);

  cardDiv.appendChild(headerDiv);
  cardDiv.appendChild(layoutDiv);
  document.body.appendChild(cardDiv);
  return layoutDiv;
}

function addDivText(div, text, extraStyles) {
  let textStyles = ["card-text"];
  if (extraStyles) {
    textStyles = textStyles.concat(extraStyles);
  }

  const textDiv = createDiv(textStyles, text.toString());
  div.appendChild(textDiv);
}

window.addEventListener('DOMContentLoaded', async () => {
  const response = await fetch('/api/values');
  const values = await response.json();

  const weatherCard = addCard("Weather", ["grid-width-4", "card-column-flex"]);
  for (let i = 0; i < values.weather.descriptions.length; i++) {
    const des = values.weather.descriptions[i];
    const icon = values.weather.icons[i];
    const temp = values.weather.temperatures[i];
    console.log(values.weather.times[i]);
    const time = new Date(values.weather.times[i]*1000);

    let timeText;
    if (time.getHours() === 0) {
      timeText = "midnight";
    } else if (time.getHours() === 12) {
      timeText = "noon";
    } else if (time.getHours() < 12) {
      timeText = `${time.getHours()}am`;
    } else {
      timeText = `${time.getHours()%12}pm`;
    }

    const iconImg = document.createElement("img")
    iconImg.src = `https://openweathermap.org/img/wn/${icon}.png`
    iconImg.width = "64"
    iconImg.height = "64"
    iconImg.alt = des;

    const reportDiv = createDiv("layer3");
    reportDiv.appendChild(iconImg)
    addDivText(reportDiv, `${temp}F`);
    addDivText(reportDiv, timeText);
    weatherCard.appendChild(reportDiv);
  }

  const changieCard = addCard("Changie", ["card-two-column"]);
  addDivText(changieCard, "Stars", "text-right");
  addDivText(changieCard, values.changie.stars);
  addDivText(changieCard, "Issues", "text-right");
  addDivText(changieCard, values.changie.issues);
  addDivText(changieCard, "Forks", "text-right");
  addDivText(changieCard, values.changie.forks);
});
