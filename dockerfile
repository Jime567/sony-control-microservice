FROM byuoitav/amd64-alpine

ARG NAME
ENV name=${NAME}

COPY ${name}-bin ${name}-bin 
COPY version.txt version.txt

# add any required files/folders here

ENTRYPOINT ./${name}-bin
