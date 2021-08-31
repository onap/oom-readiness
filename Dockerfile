FROM nexus3.onap.org:10001/onap/integration-python:7.0.0

ENV CERT="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
ENV TOKEN="/var/run/secrets/kubernetes.io/serviceaccount/token"

COPY ready.py .
COPY jobContainerReady.py .

ENTRYPOINT ["/app/ready.py"]
CMD [""]
