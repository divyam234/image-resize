# Image Resize Service

```sh
git clone https://github.com/divyam234/image-resize
cd image-resize
docker compose up -d
```
Avaliable on http://localhost:8080

Format: http://localhost:8080/{host}/{path}?w=480&h=270

Example: http://localhost:8080/images.unsplash.com/photo-1682687220067-dced9a881b56?w=480&h=270

**If any of height or width is missing image will be resize using original aspect ratio**
