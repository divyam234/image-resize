# Image Resize Service
Resize Images into Webp Format
```sh
git clone https://github.com/divyam234/image-resize
cd image-resize
docker compose up -d
```
Avaliable on http://localhost:8080

Format: http://localhost:8080/{host}/{path}?w=480&h=270&q=80

Example: http://localhost:8080/images.unsplash.com/photo-1682687220067-dced9a881b56?w=480&h=270&q=80 <br>
<br>
**If any of height or width is missing image will be resized using original aspect ratio.**
<br>
<br>
**Resizing Service works best with cache layer on top use cloudflare workers cache api ore deploy this service in front of cloudflare tunnel for caching.**
