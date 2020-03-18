FROM bash
ADD hello_world.sh /bin/hello_world.sh
RUN  chmod +x /bin/hello_world.sh
#CMD echo  '{"action_01": "data_01"}' > /workflow/data
ENTRYPOINT [ "/bin/hello_world.sh" ]
