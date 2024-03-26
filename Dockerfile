FROM nexus3.onap.org:10001/onap/integration-python:12.0.0

ENV CERT="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
ENV TOKEN="/var/run/secrets/kubernetes.io/serviceaccount/token"

COPY requirements.txt .
RUN pip install -r requirements.txt
COPY ready.py .

ENTRYPOINT ["/app/ready.py"]
CMD [""]
