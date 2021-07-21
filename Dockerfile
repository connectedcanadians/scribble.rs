FROM debian

RUN apt-get update  
RUN apt-get install -y ca-certificates

COPY ./scribblers /scribblers
ADD /templates /templates
ADD /resources /resources


CMD /scribblers