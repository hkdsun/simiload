FROM tianon/true
MAINTAINER hkdsun "hkdsun@github.com"
EXPOSE 8080
EXPOSE 8081

ADD server /

CMD ["/server"]
