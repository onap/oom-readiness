FROM registry.gitlab.com/onap-integration/docker/onap-python:frankfurt

COPY requirements.txt .

USER root
RUN pip install --no-cache-dir requirements.txt

USER onap

ENV CERT="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
ENV TOKEN="/var/run/secrets/kubernetes.io/serviceaccount/token"

COPY ready.py .

ENTRYPOINT ["/usr/onap/ready.py"]
CMD [""]
