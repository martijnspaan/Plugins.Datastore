@echo off
pushd source
call ..\Tools\go_appengine\goapp.bat serve --port=4321
popd