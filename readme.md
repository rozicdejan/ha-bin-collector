HA Bin Collector
This is a Home Assistant add-on for fetching waste collection data using a Go application. The data is retrieved from the Simbio waste collection service, and the add-on provides both a dynamic HTML page and a JSON API endpoint to access the data.

Features
Retrieves waste collection data for a configurable address.
Serves an HTML page at the root endpoint (/) to display the data.
Provides an API endpoint (/api/data) to access the data in JSON format.
Configurable through Home Assistant add-on options.
Installation
1. Place the Add-on Files in Home Assistant
Copy the bin-collector directory (containing main.go, Dockerfile, run.sh, template.html, etc.) into the Home Assistant addons directory.

2. Add-on Configuration
The add-on uses a config.json file for configuration. The address option allows you to specify the address for fetching waste collection data.

Access the API Endpoint
The add-on exposes data via a JSON API endpoint:

http://<home_assistant_ip>:8081/api/data

Example API Response
¨¨json
Copy code
{
  "name": "ZAČRET 69",
  "query": "zacret 69 ljubecna",
  "city": "LJUBEČNA",
  "mko_name": "Mešani komunalni odpadki",
  "mko_date": "petek, 15. 11. 2024",
  "emb_name": "Embalaža",
  "emb_date": "petek, 8. 11. 2024",
  "bio_name": "Biološki odpadki",
  "bio_date": "sreda, 6. 11. 2024"
}
¨¨¨
Configuration
address: Specifies the address for fetching waste collection data. This can be set in the Home Assistant add-on configuration panel.

[![Add Add-On Repository](https://my.home-assistant.io/badges/supervisor_addon_repository.svg)](https://my.home-assistant.io/redirect/supervisor_addon_repository/?repository_url=https://github.com/rozicdejan/ha-bin-collector)
