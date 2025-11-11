# Golang기반 파일 공유 서비스 개발
파일 업로드/다운로드 및 공유 URL을 통해 공유/다운로드가 가능한 서비스입니다
<br><br><br><br>
## 주요 기능
- 파일 업로드 및 다운로드<br>
- 공유 URL을 통한 파일 공유<br>
- XSS, 대용량 파일 업로드 방지, 파일 내용 검증 등 보안 적용 완료<br>
- 간단한 UI 지원<br>
<br><br>
## 기술 스택
- 언어 : Go
- 프레임워크 : net/http
- 템플릿 엔진 : html/template
- 배포 : Railway
- 파일 저장 방식 : 디렉토리('./uploads')
<br><br><br>
## 사용 예시
- 코드를 github에 push합니다
- railway, render 등을 통해 github와 연결 후 배포합니다
- 제공된 url을 통해 서비스에 접속할 수 있습니다
<br><br><br>
## 보안 참고
- 확장자가 png, pdf, jpg, gif, jpeg인 파일만 업로드 가능합니다
- 파일 크기가 64MB 이상인 파일은 업로드 불가합니다
<br><br><br>
## 라이선스
go 1.22.2
