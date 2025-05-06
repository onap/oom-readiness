FROM python:3.11-alpine

ENV CERT="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
ENV TOKEN="/var/run/secrets/kubernetes.io/serviceaccount/token"

COPY requirements.txt .
RUN pip install -r requirements.txt
COPY ready.py .

ENTRYPOINT ["/app/ready.py"]
CMD [""]
