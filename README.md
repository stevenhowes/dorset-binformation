# dorset-binformation

A data scraper for the Dorset Council bin collection site for bin collection dates

# Installation
````
go get -v github.com/stevenhowes/dorset-binformation/
make install
service dorset-binformation start
````

Edit /etc/systemd/system/dorset-binformation.service if an alternate port is required

# Example Curl
````
curl http://localhost:8998/uprn/100041115206

{"food":1712880000,"recycling":1713484800,"rubbish":1712880000}
````
