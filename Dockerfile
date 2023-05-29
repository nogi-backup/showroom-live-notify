FROM public.ecr.aws/lambda/python:3.10

WORKDIR /app
COPY requirements.txt  .
RUN  pip3 install -r requirements.txt --target /app

# Copy function code
COPY app.py /app

# Set the CMD to your handler (could also be done as a parameter override outside of the Dockerfile)
CMD [ "app.handler" ]